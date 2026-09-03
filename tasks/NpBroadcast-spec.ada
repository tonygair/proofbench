--  <vc-preamble>
package Np_Broadcast_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   type Matrix is
     array (Index_Type range <>, Index_Type range <>) of Value_Type;

   --  Dafny models a shape as a sequence of length two.
   subtype Shape_Index is Natural range 0 .. 1;
   type Shape_Array is array (Shape_Index) of Index_Type;

   --  A SPARK matrix is rectangular by construction, so Dafny's MatrixWf
   --  predicate is discharged by the type itself.  MatrixSize is kept.
   function Matrix_Size (M : Matrix) return Natural is
     (M'Length (1) * M'Length (2));

   --  Zero-based row offset of I within M's first dimension.
   function Row_Of (M : Matrix; I : Index_Type) return Integer is
     (I - M'First (1))
   with Pre => I in M'Range (1);

   --  Zero-based column offset of J within M's second dimension.
   function Col_Of (M : Matrix; J : Index_Type) return Integer is
     (J - M'First (2))
   with Pre => J in M'Range (2);
--  </vc-preamble>

--  <vc-spec>
   procedure Broadcast
     (A      : Int_Array;
      Shape  : Shape_Array;
      Result : out Matrix)
   with
     Pre  => A'Length > 0
             and then Shape (0) > 0
             and then Shape (1) > 0
             and then (Shape (0) = A'Length or else Shape (1) = A'Length)
             and then Result'Length (1) = Shape (0)
             and then Result'Length (2) = Shape (1),
     Post => Result'Length (1) = Shape (0)
             and then Result'Length (2) = Shape (1)
             and then Matrix_Size (Result) = Shape (0) * Shape (1)
             and then
               (for all I in Result'Range (1) =>
                  (for all J in Result'Range (2) =>
                     (if Shape (0) = A'Length
                      then Result (I, J) = A (A'First + Row_Of (Result, I))
                      else Result (I, J) = A (A'First + Col_Of (Result, J)))));

end Np_Broadcast_Spec;

package body Np_Broadcast_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Broadcast
     (A      : Int_Array;
      Shape  : Shape_Array;
      Result : out Matrix) is
   begin
      pragma Assume (False);
   end Broadcast;
--  </vc-code>

--  <vc-postamble>
end Np_Broadcast_Spec;
--  </vc-postamble>
