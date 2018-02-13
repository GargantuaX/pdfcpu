package validate

import (
	"github.com/hhrutter/pdfcpu/types"
	"github.com/pkg/errors"
)

func validateResourceDict(xRefTable *types.XRefTable, obj interface{}) (hasResources bool, err error) {

	logInfoValidate.Println("*** validateResourceDict begin: ***")

	dict, err := xRefTable.DereferenceDict(obj)
	if err != nil {
		return false, err
	}
	if dict == nil {
		logInfoValidate.Printf("validateResourceDict end: object  is nil.\n")
		return false, nil
	}

	for k, v := range map[string]struct {
		validate     func(xRefTable *types.XRefTable, obj interface{}, sinceVersion types.PDFVersion) error
		sinceVersion types.PDFVersion
	}{
		"ExtGState":  {validateExtGStateResourceDict, types.V10},
		"Font":       {validateFontResourceDict, types.V10},
		"XObject":    {validateXObjectResourceDict, types.V10},
		"Properties": {validatePropertiesResourceDict, types.V10},
		"ColorSpace": {validateColorSpaceResourceDict, types.V10},
		"Pattern":    {validatePatternResourceDict, types.V10},
		"Shading":    {validateShadingResourceDict, types.V13},
	} {
		if obj, ok := dict.Find(k); ok {
			err = v.validate(xRefTable, obj, v.sinceVersion)
			if err != nil {
				return false, err
			}
		}
	}

	// Beginning with PDF V1.4 this feature is considered to be obsolete.
	//_, err = validateNameArrayEntry(xRefTable, dict, "resourceDict", "ProcSet", OPTIONAL, types.V10, validateProcedureSetName)
	//if err != nil {
	//	return false, nil
	//}

	logInfoValidate.Println("*** validateResourceDict end ***")

	return true, nil
}

func validatePageContents(xRefTable *types.XRefTable, dict *types.PDFDict) (hasContents bool, err error) {

	logInfoValidate.Println("*** validatePageContents begin ***")

	obj, found := dict.Find("Contents")
	if !found {
		return false, err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil || obj == nil {
		return false, err
	}

	switch obj := obj.(type) {

	case types.PDFStreamDict:
		// no further processing.
		hasContents = true

	case types.PDFArray:
		// process array of content stream dicts.

		logInfoValidate.Printf("validatePageContents: writing content array\n")

		for _, obj := range obj {

			obj, err = xRefTable.DereferenceStreamDict(obj)
			if err != nil {
				return false, err
			}

			if obj == nil {
				continue
			}

			hasContents = true

		}

	default:
		return false, errors.Errorf("validatePageContents: page content must be stream dict or array")
	}

	logInfoValidate.Println("*** validatePageContents end ***")

	return hasContents, nil
}

func validatePageResources(xRefTable *types.XRefTable, dict *types.PDFDict, hasResources, hasContents bool) error {

	logInfoValidate.Println("*** validatePageResources begin ***")

	if obj, found := dict.Find("Resources"); found {
		_, err := validateResourceDict(xRefTable, obj)
		return err
	}

	if !hasResources && hasContents {
		return errors.New("validatePageResources: missing required entry \"Resources\" - should be inheritated")
	}

	logInfoValidate.Println("*** validatePageResources end ***")

	return nil
}

func validatePageEntryMediaBox(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) (hasMediaBox bool, err error) {

	logInfoValidate.Println("*** validatePageEntryMediaBox begin ***")

	obj, err := validateRectangleEntry(xRefTable, dict, "pageDict", "MediaBox", required, sinceVersion, nil)
	if err != nil {
		return false, err
	}
	if obj != nil {
		hasMediaBox = true
	}

	logInfoValidate.Printf("*** validatePageEntryMediaBox end: hasMediaBox=%v ***\n", hasMediaBox)

	return hasMediaBox, nil
}

func validatePageEntryCropBox(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryCropBox begin ***")

	_, err := validateRectangleEntry(xRefTable, dict, "pagesDict", "CropBox", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryCropBox end ***")

	return nil
}

func validatePageEntryBleedBox(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryBleedBox begin ***")

	_, err := validateRectangleEntry(xRefTable, dict, "pagesDict", "BleedBox", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryBleedBox end ***")

	return nil
}

func validatePageEntryTrimBox(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryTrimBox begin ***")

	_, err := validateRectangleEntry(xRefTable, dict, "pagesDict", "TrimBox", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryTrimBox end ***")

	return nil
}

func validatePageEntryArtBox(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryArtBox begin ***")

	_, err := validateRectangleEntry(xRefTable, dict, "pagesDict", "ArtBox", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryArtBox end ***")

	return nil
}

func validateBoxStyleDictEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion) (*types.PDFDict, error) {

	logInfoValidate.Println("*** validateBoxStyleDictEntry begin ***")

	d, err := validateDictEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, nil)
	if err != nil {
		return nil, err
	}
	if d == nil {
		logInfoValidate.Printf("validateBoxStyleDictEntry end: is nil.\n")
		return nil, nil
	}

	dictName = "boxStyleDict"

	// C, number array with 3 elements, optional
	_, err = validateNumberArrayEntry(xRefTable, d, dictName, "C", OPTIONAL, sinceVersion, func(arr types.PDFArray) bool { return len(arr) == 3 })
	if err != nil {
		return nil, err
	}

	// W, number, optional
	_, err = validateNumberEntry(xRefTable, d, dictName, "W", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return nil, err
	}

	// S, name, optional
	_, err = validateNameEntry(xRefTable, d, dictName, "S", OPTIONAL, sinceVersion, validateGuideLineStyle)
	if err != nil {
		return nil, err
	}

	// D, array, optional, since V1.3, dashArray
	_, err = validateIntegerArrayEntry(xRefTable, d, dictName, "D", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return nil, err
	}

	logInfoValidate.Println("*** validateBoxStyleDictEntry end ***")

	return d, err
}

func validatePageBoxColorInfo(xRefTable *types.XRefTable, pageDict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	// box color information dict
	// see 14.11.2.2

	logInfoValidate.Println("*** validatePageBoxColorInfo begin ***")

	var dict *types.PDFDict

	dictName := "pageDict"

	dict, err := validateDictEntry(xRefTable, pageDict, dictName, "BoxColorInfo", required, sinceVersion, nil)
	if err != nil || dict == nil {
		return err
	}

	dictName = "boxColorInfoDict"

	_, err = validateBoxStyleDictEntry(xRefTable, dict, dictName, "CropBox", OPTIONAL, sinceVersion)
	if err != nil {
		return err
	}

	_, err = validateBoxStyleDictEntry(xRefTable, dict, dictName, "BleedBox", OPTIONAL, sinceVersion)
	if err != nil {
		return err
	}

	_, err = validateBoxStyleDictEntry(xRefTable, dict, dictName, "TrimBox", OPTIONAL, sinceVersion)
	if err != nil {
		return err
	}

	_, err = validateBoxStyleDictEntry(xRefTable, dict, dictName, "ArtBox", OPTIONAL, sinceVersion)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageBoxColorInfo end ***")

	return nil
}

func validatePageEntryRotate(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryRotate begin ***")

	_, err := validateIntegerEntry(xRefTable, dict, "pagesDict", "Rotate", required, sinceVersion, validateRotate)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryRotate end ***")

	return nil
}

func validatePageEntryGroup(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryGroup begin ***")

	d, err := validateDictEntry(xRefTable, dict, "pageDict", "Group", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	if d != nil {
		err = validateGroupAttributesDict(xRefTable, *d)
		if err != nil {
			return err
		}
	}

	logInfoValidate.Println("*** validatePageEntryGroup end ***")

	return nil
}

func validatePageEntryThumb(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryThumb begin ***")

	streamDict, err := validateStreamDictEntry(xRefTable, dict, "pagesDict", "Thumb", required, sinceVersion, nil)
	if err != nil {
		return err
	}
	if streamDict == nil {
		logInfoValidate.Println("validatePageEntryThumb end: is nil.")
		return nil
	}

	err = validateXObjectStreamDict(xRefTable, streamDict)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryThumb end ***")

	return nil
}

func validatePageEntryB(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	// Note: Only makes sense if "Threads" entry in document root and bead dicts present.

	logInfoValidate.Println("*** validatePageEntryB begin ***")

	arr, err := validateIndRefArrayEntry(xRefTable, dict, "pagesDict", "B", required, sinceVersion, nil)
	if err != nil {
		return err
	}
	if arr == nil {
		logInfoValidate.Println("validatePageEntryB end: is nil.")
		return nil
	}

	logInfoValidate.Println("*** validatePageEntryB end ***")

	return nil
}

func validatePageEntryDur(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("*** validatePageEntryDur begin ***")

	_, err := validateNumberEntry(xRefTable, dict, "pagesDict", "Dur", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryDur end ***")

	return nil
}

func validateTransitionDictEntryDi(xRefTable *types.XRefTable, dict *types.PDFDict) error {

	obj, found := dict.Find("Di")
	if !found {
		return nil
	}

	switch obj := obj.(type) {

	case types.PDFInteger:
		if !validateDi(obj.Value()) {
			return errors.New("validateTransitionDict: entry Di int value undefined")
		}

	case types.PDFName:
		if obj.Value() != "None" {
			return errors.New("validateTransitionDict: entry Di name value undefined")
		}
	}

	return nil
}

func validateTransitionDictEntryM(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, transStyle *types.PDFName) error {

	validateM := func(s string) bool {
		return validateTransitionDirectionOfMotion(s) &&
			(transStyle != nil && (*transStyle == "Split" || *transStyle == "Box" || *transStyle == "Fly"))
	}

	_, err := validateNameEntry(xRefTable, dict, dictName, "M", OPTIONAL, types.V10, validateM)

	return err
}

func validateTransitionDict(xRefTable *types.XRefTable, dict *types.PDFDict) error {

	logInfoValidate.Println("*** validateTransitionDict begin ***")

	dictName := "transitionDict"

	// S, name, optional
	validate := validateTransitionStyle
	if xRefTable.Version() >= types.V15 {
		validate = validateTransitionStyleV15
	}
	transStyle, err := validateNameEntry(xRefTable, dict, dictName, "S", OPTIONAL, types.V10, validate)
	if err != nil {
		return err
	}

	// D, optional, number > 0
	_, err = validateNumberEntry(xRefTable, dict, dictName, "D", OPTIONAL, types.V10, func(f float64) bool { return f > 0 })
	if err != nil {
		return err
	}

	// Dm, optional, name
	validateDm := func(s string) bool {
		return validateTransitionDimension(s) && (transStyle != nil && (*transStyle == "Split" || *transStyle == "Blinds"))
	}
	_, err = validateNameEntry(xRefTable, dict, dictName, "Dm", OPTIONAL, types.V10, validateDm)
	if err != nil {
		return err
	}

	// M, optional, name
	err = validateTransitionDictEntryM(xRefTable, dict, dictName, transStyle)
	if err != nil {
		return err
	}

	// Di, optional, number or name
	err = validateTransitionDictEntryDi(xRefTable, dict)
	if err != nil {
		return err
	}

	// SS, optional, number, since V1.5
	if transStyle != nil && *transStyle == "Fly" {
		_, err = validateNumberEntry(xRefTable, dict, dictName, "SS", OPTIONAL, types.V15, nil)
		if err != nil {
			return err
		}
	}

	// B, optional, boolean, since V1.5
	validateB := func(b bool) bool { return transStyle != nil && *transStyle == "Fly" }
	_, err = validateBooleanEntry(xRefTable, dict, dictName, "B", OPTIONAL, types.V15, validateB)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validateTransitionDict end ***")

	return nil
}

func validatePageEntryTrans(xRefTable *types.XRefTable, pageDict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryTrans begin ***")

	dict, err := validateDictEntry(xRefTable, pageDict, "pagesDict", "Trans", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	if dict == nil {
		logInfoValidate.Println("validatePageEntryTrans end: is nil.")
		return err
	}

	err = validateTransitionDict(xRefTable, dict)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryTrans begin ***")

	return nil
}

func validatePageEntryStructParents(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("*** validatePageEntryStructParents begin ***")

	_, err := validateIntegerEntry(xRefTable, dict, "pagesDict", "StructParents", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryStructParents end ***")

	return nil
}

func validatePageEntryID(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryID begin ***")

	_, err := validateStringEntry(xRefTable, dict, "pagesDict", "ID", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryID end ***")

	return nil
}

func validatePageEntryPZ(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	// Preferred zoom factor, number

	logInfoValidate.Println("*** validatePageEntryPZ begin ***")

	_, err := validateNumberEntry(xRefTable, dict, "pagesDict", "PZ", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryPZ end ***")

	return nil
}

func validatePageEntrySeparationInfo(xRefTable *types.XRefTable, pagesDict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	// see 14.11.4

	logInfoValidate.Printf("*** validatePageEntrySeparationInfo begin ***")

	dict, err := validateDictEntry(xRefTable, pagesDict, "pagesDict", "SeparationInfo", required, sinceVersion, nil)
	if err != nil || dict == nil {
		return err
	}

	dictName := "separationDict"

	_, err = validateIndRefArrayEntry(xRefTable, dict, dictName, "Pages", REQUIRED, sinceVersion, nil)
	if err != nil {
		return err
	}

	err = validateNameOrStringEntry(xRefTable, dict, dictName, "DeviceColorant", required, sinceVersion)
	if err != nil {
		return err
	}

	arr, err := validateArrayEntry(xRefTable, dict, dictName, "ColorSpace", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}
	if arr != nil {
		err = validateColorSpaceArraySubset(xRefTable, arr, []string{"Separation", "DeviceN"})
		if err != nil {
			return err
		}
	}

	logInfoValidate.Println("*** validatePageEntrySeparationInfo end ***")

	return nil
}

func validatePageEntryTabs(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryTabs begin ***")

	// Include out of spec entry "W"
	validateTabs := func(s string) bool { return memberOf(s, []string{"R", "C", "S", "W"}) }

	_, err := validateNameEntry(xRefTable, dict, "pagesDict", "Tabs", required, sinceVersion, validateTabs)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryTabs end ***")

	return nil
}

func validatePageEntryTemplateInstantiated(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	// see 12.7.6

	logInfoValidate.Println("*** validatePageEntryTemplateInstantiated begin ***")

	_, err := validateNameEntry(xRefTable, dict, "pagesDict", "TemplateInstantiated", required, sinceVersion, nil)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryTemplateInstantiated end ***")

	return nil
}

// TODO implement
func validatePageEntryPresSteps(xRefTable *types.XRefTable, pagesDict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	// see 12.4.4.2

	logInfoValidate.Println("*** validatePageEntryPresSteps begin ***")

	dict, err := validateDictEntry(xRefTable, pagesDict, "pagesDict", "PresSteps", required, sinceVersion, nil)
	if err != nil {
		return err
	}
	if dict == nil {
		logInfoValidate.Println("validatePageEntryPresSteps end: is nil.")
		return nil
	}

	return errors.New("*** validatePageEntryPresSteps: not supported ***")
}

func validatePageEntryUserUnit(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validatePageEntryUserUnit begin ***")

	// UserUnit, optional, positive number, since V1.6
	_, err := validateNumberEntry(xRefTable, dict, "pagesDict", "UserUnit", required, sinceVersion, func(f float64) bool { return f > 0 })
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validatePageEntryUserUnit end ***")

	return nil
}

func validateNumberFormatDict(xRefTable *types.XRefTable, dict *types.PDFDict, sinceVersion types.PDFVersion) error {

	dictName := "numberFormatDict"

	// Type, name, optional
	_, err := validateNameEntry(xRefTable, dict, dictName, "Type", OPTIONAL, sinceVersion, func(s string) bool { return s == "NumberFormat" })
	if err != nil {
		return err
	}

	// U, text string, required
	_, err = validateStringEntry(xRefTable, dict, dictName, "U", REQUIRED, sinceVersion, nil)
	if err != nil {
		return err
	}

	// C, number, required
	_, err = validateNumberEntry(xRefTable, dict, dictName, "C", REQUIRED, sinceVersion, nil)
	if err != nil {
		return err
	}

	// F, name, optional
	_, err = validateNameEntry(xRefTable, dict, dictName, "F", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	// D, integer, optional
	_, err = validateIntegerEntry(xRefTable, dict, dictName, "D", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	// FD, bool, optional
	_, err = validateBooleanEntry(xRefTable, dict, dictName, "FD", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	// RT, text string, optional
	_, err = validateStringEntry(xRefTable, dict, dictName, "RT", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	// RD, text string, optional
	_, err = validateStringEntry(xRefTable, dict, dictName, "RD", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	// PS, text string, optional
	_, err = validateStringEntry(xRefTable, dict, dictName, "PS", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	// SS, text string, optional
	_, err = validateStringEntry(xRefTable, dict, dictName, "SS", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	// O, name, optional
	_, err = validateNameEntry(xRefTable, dict, dictName, "O", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	return nil
}

func validateNumberFormatArrayEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	arr, err := validateArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, nil)
	if err != nil {
		return err
	}

	if arr == nil {
		return nil
	}

	for _, v := range *arr {

		d, err := xRefTable.DereferenceDict(v)
		if err != nil {
			return err
		}

		if d == nil {
			continue
		}

		err = validateNumberFormatDict(xRefTable, d, sinceVersion)
		if err != nil {
			return err
		}

	}

	return nil
}

func validateMeasureDict(xRefTable *types.XRefTable, dict *types.PDFDict, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validateMeasureDict begin ***")

	dictName := "measureDict"

	_, err := validateNameEntry(xRefTable, dict, dictName, "Type", OPTIONAL, sinceVersion, func(s string) bool { return s == "Measure" })
	if err != nil {
		return err
	}

	_, err = validateNameEntry(xRefTable, dict, dictName, "Subtype", OPTIONAL, sinceVersion, func(s string) bool { return s == "RL" })
	if err != nil {
		return err
	}

	// R, text string, required, scale ratio
	_, err = validateStringEntry(xRefTable, dict, dictName, "R", REQUIRED, sinceVersion, nil)
	if err != nil {
		return err
	}

	// X, number format array, required, for measurement of change along the x axis and, if Y is not present, along the y axis as well.
	err = validateNumberFormatArrayEntry(xRefTable, dict, dictName, "X", REQUIRED, sinceVersion)
	if err != nil {
		return err
	}

	// Y, number format array, required when the x and y scales have different units or conversion factors.
	err = validateNumberFormatArrayEntry(xRefTable, dict, dictName, "Y", OPTIONAL, sinceVersion)
	if err != nil {
		return err
	}

	// D, number format array, required, for measurement of distance in any direction.
	err = validateNumberFormatArrayEntry(xRefTable, dict, dictName, "D", REQUIRED, sinceVersion)
	if err != nil {
		return err
	}

	// A, number format array, required, for measurement of area.
	err = validateNumberFormatArrayEntry(xRefTable, dict, dictName, "A", REQUIRED, sinceVersion)
	if err != nil {
		return err
	}

	// T, number format array, optional, for measurement of angles.
	err = validateNumberFormatArrayEntry(xRefTable, dict, dictName, "T", OPTIONAL, sinceVersion)
	if err != nil {
		return err
	}

	// S, number format array, optional, for fmeasurement of the slope of a line.
	err = validateNumberFormatArrayEntry(xRefTable, dict, dictName, "S", OPTIONAL, sinceVersion)
	if err != nil {
		return err
	}

	// O, number array, optional, array of two numbers that shall specify the origin of the measurement coordinate system in default user space coordinates.
	_, err = validateNumberArrayEntry(xRefTable, dict, dictName, "O", OPTIONAL, sinceVersion, func(a types.PDFArray) bool { return len(a) == 2 })
	if err != nil {
		return err
	}

	// CYX, number, optional, a factor that shall be used to convert the largest units along the y axis to the largest units along the x axis.
	_, err = validateNumberEntry(xRefTable, dict, dictName, "CYX", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	logInfoValidate.Println("*** validateMeasureDict end ***")

	return nil
}

func validateViewportDict(xRefTable *types.XRefTable, dict *types.PDFDict, sinceVersion types.PDFVersion) error {

	logInfoValidate.Println("*** validateViewportDict begin ***")

	dictName := "viewportDict"

	_, err := validateNameEntry(xRefTable, dict, dictName, "Type", OPTIONAL, sinceVersion, func(s string) bool { return s == "Viewport" })
	if err != nil {
		return err
	}

	_, err = validateRectangleEntry(xRefTable, dict, dictName, "BBox", REQUIRED, sinceVersion, nil)
	if err != nil {
		return err
	}

	_, err = validateStringEntry(xRefTable, dict, dictName, "Name", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	// Measure, optional, dict
	d, err := validateDictEntry(xRefTable, dict, dictName, "Measure", OPTIONAL, sinceVersion, nil)
	if err != nil {
		return err
	}

	if d != nil {
		err = validateMeasureDict(xRefTable, d, sinceVersion)
		if err != nil {
			return err
		}
	}

	logInfoValidate.Println("*** validateViewportDict end ***")

	return nil
}

func validatePageEntryVP(xRefTable *types.XRefTable, pagesDict *types.PDFDict, required bool, sinceVersion types.PDFVersion) error {

	// see table 260

	logInfoValidate.Println("*** validatePageEntryVP begin ***")

	arr, err := validateArrayEntry(xRefTable, pagesDict, "pagesDict", "VP", required, sinceVersion, nil)
	if err != nil {
		return err
	}
	if arr == nil {
		logInfoValidate.Println("validatePageEntryVP end: is nil.")
		return nil
	}

	var dict *types.PDFDict

	for _, v := range *arr {

		if v == nil {
			continue
		}

		dict, err = xRefTable.DereferenceDict(v)
		if err != nil {
			return err
		}

		if dict == nil {
			continue
		}

		err = validateViewportDict(xRefTable, dict, sinceVersion)
		if err != nil {
			return err
		}

	}

	logInfoValidate.Printf("*** validatePageEntryVP end ***")

	return nil
}

func validatePageDict(xRefTable *types.XRefTable, pageDict *types.PDFDict, objNumber, genNumber int, hasResources, hasMediaBox bool) error {

	logInfoValidate.Printf("*** validatePageDict begin: hasResources=%v hasMediaBox=%v obj#%d ***\n", hasResources, hasMediaBox, objNumber)

	if indref := pageDict.IndirectRefEntry("Parent"); indref == nil {
		return errors.New("validatePageDict: missing parent")
	}

	// Contents
	hasContents, err := validatePageContents(xRefTable, pageDict)
	if err != nil {
		return err
	}

	// Resources
	err = validatePageResources(xRefTable, pageDict, hasResources, hasContents)
	if err != nil {
		return err
	}

	// MediaBox
	_, err = validatePageEntryMediaBox(xRefTable, pageDict, !hasMediaBox, types.V10)
	if err != nil {
		return err
	}

	// PieceInfo
	sinceVersion := types.V13
	if xRefTable.ValidationMode == types.ValidationRelaxed {
		sinceVersion = types.V10
	}
	hasPieceInfo, err := validatePieceInfo(xRefTable, pageDict, OPTIONAL, sinceVersion)
	if err != nil {
		return err
	}

	// LastModified
	lm, err := validateDateEntry(xRefTable, pageDict, "pageDict", "LastModified", OPTIONAL, types.V13)
	if err != nil {
		return err
	}

	if hasPieceInfo && lm == nil && xRefTable.ValidationMode == types.ValidationStrict {
		return errors.New("validatePageDict: missing \"LastModified\" (required by \"PieceInfo\")")
	}

	// AA
	err = validateAdditionalActions(xRefTable, pageDict, "pageDict", "AA", OPTIONAL, types.V14, "page")
	if err != nil {
		return err
	}

	type v struct {
		validate     func(xRefTable *types.XRefTable, dict *types.PDFDict, required bool, sinceVersion types.PDFVersion) (err error)
		required     bool
		sinceVersion types.PDFVersion
	}

	for _, f := range []v{
		{validatePageEntryCropBox, OPTIONAL, types.V10},
		{validatePageEntryBleedBox, OPTIONAL, types.V13},
		{validatePageEntryTrimBox, OPTIONAL, types.V13},
		{validatePageEntryArtBox, OPTIONAL, types.V13},
		{validatePageBoxColorInfo, OPTIONAL, types.V14},
		{validatePageEntryRotate, OPTIONAL, types.V10},
		{validatePageEntryGroup, OPTIONAL, types.V14},
		{validatePageEntryThumb, OPTIONAL, types.V10},
		{validatePageEntryB, OPTIONAL, types.V11},
		{validatePageEntryDur, OPTIONAL, types.V11},
		{validatePageEntryTrans, OPTIONAL, types.V11},
		{validateMetadata, OPTIONAL, types.V14},
		{validatePageEntryStructParents, OPTIONAL, types.V10},
		{validatePageEntryID, OPTIONAL, types.V13},
		{validatePageEntryPZ, OPTIONAL, types.V13},
		{validatePageEntrySeparationInfo, OPTIONAL, types.V13},
		{validatePageEntryTabs, OPTIONAL, types.V15},
		{validatePageEntryTemplateInstantiated, OPTIONAL, types.V15},
		{validatePageEntryPresSteps, OPTIONAL, types.V15},
		{validatePageEntryUserUnit, OPTIONAL, types.V16},
		{validatePageEntryVP, OPTIONAL, types.V16},
	} {
		err = f.validate(xRefTable, pageDict, f.required, f.sinceVersion)
		if err != nil {
			return err
		}
	}

	logInfoValidate.Printf("*** validatePageDict end: obj#%d ***\n", objNumber)

	return nil
}

func validatePagesDictGeneralEntries(xRefTable *types.XRefTable, dict *types.PDFDict) (hasResources, hasMediaBox bool, err error) {

	hasResources, err = validateResources(xRefTable, dict)
	if err != nil {
		return false, false, err
	}

	// MediaBox: optional, rectangle
	hasMediaBox, err = validatePageEntryMediaBox(xRefTable, dict, OPTIONAL, types.V10)
	if err != nil {
		return false, false, err
	}

	// CropBox: optional, rectangle
	err = validatePageEntryCropBox(xRefTable, dict, OPTIONAL, types.V10)
	if err != nil {
		return false, false, err
	}

	// Rotate:  optional, integer
	err = validatePageEntryRotate(xRefTable, dict, OPTIONAL, types.V10)
	if err != nil {
		return false, false, err
	}

	return hasResources, hasMediaBox, nil
}

func dictTypeForPageNodeDict(pageNodeDict *types.PDFDict) (string, error) {

	if pageNodeDict == nil {
		return "", errors.New("dictTypeForPageNodeDict: pageNodeDict is null")
	}

	dictType := pageNodeDict.Type()
	if dictType == nil {
		return "", errors.New("dictTypeForPageNodeDict: missing pageNodeDict type")
	}

	return *dictType, nil
}

func validateResources(xRefTable *types.XRefTable, dict *types.PDFDict) (hasResources bool, err error) {

	// Get number of pages of this PDF file.
	pageCount := dict.IntEntry("Count")
	if pageCount == nil {
		return false, errors.New("validateResources: missing \"Count\"")
	}

	// TODO not ideal - overall pageCount is only set during validation!
	if xRefTable.PageCount == 0 {
		xRefTable.PageCount = *pageCount
	}

	logInfoValidate.Printf("validateResources: This page node has %d pages\n", *pageCount)

	// Resources: optional, dict
	obj, ok := dict.Find("Resources")
	if !ok {
		return false, nil
	}

	return validateResourceDict(xRefTable, obj)
}

func validatePagesDict(xRefTable *types.XRefTable, dict *types.PDFDict, objNumber, genNumber int, hasResources, hasMediaBox bool) error {

	logInfoValidate.Printf("*** validatePagesDict begin: hasResources=%v hasMediaBox=%v obj#%d ***\n", hasResources, hasMediaBox, objNumber)

	// Resources and Mediabox are inheritated.
	//var dHasResources, dHasMediaBox bool
	dHasResources, dHasMediaBox, err := validatePagesDictGeneralEntries(xRefTable, dict)
	if err != nil {
		return err
	}

	if dHasResources {
		hasResources = true
	}

	if dHasMediaBox {
		hasMediaBox = true
	}

	// Iterate over page tree.
	kidsArray := dict.PDFArrayEntry("Kids")
	if kidsArray == nil {
		return errors.New("validatePagesDict: corrupt \"Kids\" entry")
	}

	for _, obj := range *kidsArray {

		if obj == nil {
			logDebugValidate.Println("validatePagesDict: kid is nil")
			continue
		}

		// Dereference next page node dict.
		indRef, ok := obj.(types.PDFIndirectRef)
		if !ok {
			return errors.New("validatePagesDict: missing indirect reference for kid")
		}

		logInfoValidate.Printf("validatePagesDict: PageNode: %s\n", indRef)

		objNumber := indRef.ObjectNumber.Value()
		genNumber := indRef.GenerationNumber.Value()

		var pageNodeDict *types.PDFDict
		pageNodeDict, err = xRefTable.DereferenceDict(indRef)
		if err != nil {
			return err
		}

		dictType, err := dictTypeForPageNodeDict(pageNodeDict)
		if err != nil {
			return err
		}

		switch dictType {

		case "Pages":
			// Recurse over pagetree
			err = validatePagesDict(xRefTable, pageNodeDict, objNumber, genNumber, hasResources, hasMediaBox)
			if err != nil {
				return err
			}

		case "Page":
			err = validatePageDict(xRefTable, pageNodeDict, objNumber, genNumber, hasResources, hasMediaBox)
			if err != nil {
				return err
			}

		default:
			return errors.Errorf("validatePagesDict: Unexpected dict type: %s", dictType)

		}

	}

	logInfoValidate.Printf("*** validatePagesDict end: obj#%d ***\n", objNumber)

	return nil
}

func validatePages(xRefTable *types.XRefTable, rootDict *types.PDFDict) (*types.PDFDict, error) {

	logInfoValidate.Println("*** validatePages begin: ***")

	// Ensure indirect reference entry "Pages".
	indRef := rootDict.IndirectRefEntry("Pages")
	if indRef == nil {
		return nil, errors.New("validatePages: missing indirect obj for pages dict")
	}

	objNumber := indRef.ObjectNumber.Value()
	genNumber := indRef.GenerationNumber.Value()

	// Dereference root of page node tree.
	rootPageNodeDict, err := xRefTable.DereferenceDict(*indRef)
	if err != nil {
		return nil, err
	}

	if rootPageNodeDict == nil {
		return nil, errors.New("validatePagesDict: cannot dereference pageNodeDict")
	}

	// Process page node tree.
	err = validatePagesDict(xRefTable, rootPageNodeDict, objNumber, genNumber, false, false)
	if err != nil {
		return nil, err
	}

	logInfoValidate.Println("*** validatePages end: ***")

	return rootPageNodeDict, nil
}